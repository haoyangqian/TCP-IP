
def makeLinkFile(src, neighbors) 
	filename = "#{src}.lnx"
	service = "localhost:600#{src}"

	failed = false	

	File.open(filename, "w") do |f|
		begin
			f.puts(service)

			neighbors.each do |n|
				neighbor_service = "localhost:600#{n}"

				#(1..2).each do |i|
				i = 1
					src_link = "#{src}.#{n}.0.#{i}"
					dst_link = "#{n}.#{src}.0.#{i}"
					f.puts("#{neighbor_service} #{src_link} #{dst_link}")
				# end
			end
		rescue
			failed = true
		end
	end

	File.delete(filename) if failed
end	 


makeLinkFile(1, [2, 5])
makeLinkFile(2, [1, 3, 6])
makeLinkFile(3, [2, 4, 7])
makeLinkFile(4, [3, 8])
makeLinkFile(5, [1, 6])
makeLinkFile(6, [2, 5, 7])
makeLinkFile(7, [3, 6, 8])
makeLinkFile(8, [4, 7])
