
def makeLinkFile(src, neighbors, link_counts)
	filename = "#{src}.lnx"
    port = 6000 + src
	service = "localhost:#{port}"

	failed = false

	File.open(filename, "w") do |f|
		begin
			f.puts(service)

			neighbors.each do |n|
                neighbor_port = 6000 + n
				neighbor_service = "localhost:#{neighbor_port}"

				(1..link_counts).each do |i|
					src_link = "#{src}.#{n}.0.#{i}"
					dst_link = "#{n}.#{src}.0.#{i}"
					f.puts("#{neighbor_service} #{src_link} #{dst_link}")
				end
			end
		rescue
			failed = true
		end
	end

	File.delete(filename) if failed
end


makeLinkFile(1, [2, 5], 3)
makeLinkFile(2, [1, 3, 6], 3)
makeLinkFile(3, [2, 4, 7], 3)
makeLinkFile(4, [3, 8], 3)
makeLinkFile(5, [1, 6], 3)
makeLinkFile(6, [2, 5, 7],  3)
makeLinkFile(7, [3, 6, 8], 3)
makeLinkFile(8, [4, 7], 3)
