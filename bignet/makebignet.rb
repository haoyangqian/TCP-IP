
def makeLinkFile(src, neighbors) 
	filename = "bignet-#{src}.lnx"
	service = "localhost:600#{src}"

	failed = false	

	File.open(filename, "w") do |f|
		begin
			f.puts(service)

			neighbors.each do |n|
				neighbor_service = "localhost:600#{n}"

				(1..2).each do |i|
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


makeLinkFile(1, [2, 4])
makeLinkFile(2, [1, 3, 5])
makeLinkFile(3, [2, 6])
makeLinkFile(4, [1, 5, 7])
makeLinkFile(5, [2, 4, 6, 8])
makeLinkFile(6, [3, 5, 9])
makeLinkFile(7, [4, 8])
makeLinkFile(8, [5, 7, 9])
makeLinkFile(9, [6, 8])